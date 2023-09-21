// MIT License
//
// Copyright (c) 2020 Dmitrii Ustiugov and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
        "bufio"
        "context"
        "encoding/json"
        "flag"
        "fmt"
        "io/ioutil"
        "os"
        "strconv"
        "sync"
        "sync/atomic"
        "time"

        ctrdlog "github.com/containerd/containerd/log"
        "github.com/google/uuid"
        log "github.com/sirupsen/logrus"
        "google.golang.org/grpc"

        "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

        "github.com/vhive-serverless/vhive/utils/benchmarking/eventing/vhivemetadata"

        "github.com/vhive-serverless/vhive/examples/endpoint"
        tracing "github.com/vhive-serverless/vhive/utils/tracing/go"
)

const TimeseriesDBAddr = "10.96.0.84:90"

var (
        completed   int64
        latSlice    LatencySlice
        portFlag    *int
        grpcTimeout time.Duration
        withTracing *bool
        workflowIDs map[*endpoint.Endpoint]string
)

func main() {
        endpointsFile := flag.String("endpointsFile", "endpoints.json", "File with endpoints' metadata")
        allReq := flag.Int("allreq", 1, "All requests sent")
        waitDuration := flag.Int("wait", 10, "Time to wait requests completed")
        // rps := flag.Float64("rps", 1.0, "Target requests per second")
        runDuration := flag.Int("time", 1, "Send all requests at a constant rate within X seconds")
        latencyOutputFile := flag.String("latf", "lat.csv", "CSV file for the latency measurements in microseconds")
        portFlag = flag.Int("port", 80, "The port that functions listen to")
        withTracing = flag.Bool("trace", false, "Enable tracing in the client")
        zipkin := flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
        debug := flag.Bool("dbg", false, "Enable debug logging")
        grpcTimeout = time.Duration(*flag.Int("grpcTimeout", 30, "Timeout in seconds for gRPC requests")) * time.Second

        flag.Parse()

        log.SetFormatter(&log.TextFormatter{
                TimestampFormat: ctrdlog.RFC3339NanoFixed,
                FullTimestamp:   true,
        })
        log.SetOutput(os.Stdout)
        if *debug {
                log.SetLevel(log.DebugLevel)
                log.Debug("Debug logging is enabled")
        } else {
                log.SetLevel(log.InfoLevel)
        }

        log.Info("Reading the endpoints from the file: ", *endpointsFile)

        endpoints, err := readEndpoints(*endpointsFile)
        if err != nil {
                log.Fatal("Failed to read the endpoints file: ", err)
        }

        workflowIDs = make(map[*endpoint.Endpoint]string)
        for _, ep := range endpoints {
                workflowIDs[ep] = uuid.New().String()
        }

        if *withTracing {
                shutdown, err := tracing.InitBasicTracer(*zipkin, "invoker")
                if err != nil {
                        log.Print(err)
                }
                defer shutdown()
        }

        realRPS := runExperiment(endpoints, *runDuration, *allReq, *waitDuration)

        writeLatencies(realRPS, *latencyOutputFile)
}

func readEndpoints(path string) (endpoints []*endpoint.Endpoint, _ error) {
        data, err := ioutil.ReadFile(path)
        if err != nil {
                return nil, err
        }
        if err := json.Unmarshal(data, &endpoints); err != nil {
                return nil, err
        }
        return
}

func runExperiment(endpoints []*endpoint.Endpoint, runDuration int, allReq int, waitDuration int) (realRPS float64) {
        var issued int

        Start(TimeseriesDBAddr, endpoints, workflowIDs)

        //timeout := time.After(time.Duration(runDuration) * time.Second)
        wait := time.After(time.Duration(waitDuration) * time.Second)
        //tick := time.Tick(time.Duration(1000) * time.Millisecond)
        start := time.Now()

        for issued < allReq {
                for _, ep := range endpoints {
                        for i := 0; i < allReq / runDuration; i++ {
                                if ep.Eventing {
                                        go invokeEventingFunction(ep)
                                } else {
                                        go invokeServingFunction(ep)
                                }
                        }
                }
                issued += allReq / runDuration
                time.Sleep(time.Second)
        }
// loop:
//      for {
//              select {
//              case <- timeout:
//                      break loop
//              case <- tick:

//                      continue
//              }
//      }

        <- wait
        duration := time.Since(start).Seconds()
        realRPS = float64(completed) / duration
        addDurations("end", End())
        log.Infof("Issued / completed requests: %d, %d", issued, completed)
        log.Infof("Real / target RPS: %.2f / %v", realRPS, 1)
        log.Println("Experiment finished!")
        return
}

func SayHello(address, workflowID string) {
        dialOptions := []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
        if *withTracing {
                dialOptions = append(dialOptions, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
        }
        conn, err := grpc.Dial(address, dialOptions...)
        if err != nil {
                log.Fatalf("did not connect: %v", err)
        }
        defer conn.Close()

        c := NewGreeterClient(conn)

        ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
        defer cancel()

        _, err = c.SayHello(ctx, &HelloRequest{
                Name: "faas",
                VHiveMetadata: vhivemetadata.MakeVHiveMetadata(
                        workflowID,
                        uuid.New().String(),
                        time.Now().UTC(),
                ),
        })
        if err != nil {
                log.Warnf("Failed to invoke %v, err=%v", address, err)
        } else {
                atomic.AddInt64(&completed, 1)
        }
}

func invokeEventingFunction(endpoint *endpoint.Endpoint) {
        address := fmt.Sprintf("%s:%d", endpoint.Hostname, *portFlag)
        log.Debug("Invoking asynchronously: ", address)

        SayHello(address, workflowIDs[endpoint])
}

func invokeServingFunction(endpoint *endpoint.Endpoint) {
        defer getDuration(startMeasurement(endpoint.Hostname)) // measure entire invocation time

        address := fmt.Sprintf("%s:%d", endpoint.Hostname, *portFlag)
        log.Debug("Invoking: ", address)

        SayHello(address, workflowIDs[endpoint])
}

// LatencySlice is a thread-safe slice to hold a slice of latency measurements.
type LatencySlice struct {
        sync.Mutex
        function [] string
        slice []int64
}

func startMeasurement(msg string) (string, time.Time) {
        return msg, time.Now()
}

func getDuration(msg string, start time.Time) {
        latency := time.Since(start)
        log.Debugf("Invoked %v in %v usec\n", msg, latency.Microseconds())
        addDurations(msg, []time.Duration{latency})
}

func addDurations(msg string, ds []time.Duration) {
        latSlice.Lock()
        for _, d := range ds {
                latSlice.slice = append(latSlice.slice, d.Microseconds())
                latSlice.function = append(latSlice.function, msg)
        }
        latSlice.Unlock()
}

func writeLatencies(rps float64, latencyOutputFile string) {
        latSlice.Lock()
        defer latSlice.Unlock()

        fileName := fmt.Sprintf("rps%.2f_%s", rps, latencyOutputFile)
        log.Info("The measured latencies are saved in ", fileName)

        file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

        if err != nil {
                log.Fatal("Failed creating file: ", err)
        }

        datawriter := bufio.NewWriter(file)

        for i, lat := range latSlice.slice {
                _, err := datawriter.WriteString(latSlice.function[i] + ", " + strconv.FormatInt(lat, 10) + "\n")
                if err != nil {
                        log.Fatal("Failed to write the URLs to a file ", err)
                }
        }

        datawriter.Flush()
        file.Close()
}