import csv

for i in [1]:
    print("helloworld" + str(i))
    result_range = []
    result_average = []
    result_max = []
    result_min = []
    for j in [1,2,4,8,16,32,64]:
        values = []
        for k in range(1,6):
            filename = "helloworld.json_" + str(i) + "_" + str(j) + "_" + str(k) + ".csv"
            with open(filename) as csvfile:
                csv_reader = csv.reader(csvfile)
                value = 0
                for row in csv_reader:
                    value += int(row[1])
            values.append(value/j)
        values.sort()
        dis = values[-1] - values[0]
        result_range.append(dis / values[0])
        result_average.append(sum(values) / len(values))
        result_max.append(values[-1])
        result_min.append(values[0])

    print("The max-min range is")
    print(" ".join(str(c*100) for c in result_range))
    print("The average latency is")
    print(" ".join(str(c/1000000) for c in result_average))
    print("The max latency is")
    print(" ".join(str(c/1000000) for c in result_max))
    print("The min latency is")
    print(" ".join(str(c/1000000) for c in result_min))

