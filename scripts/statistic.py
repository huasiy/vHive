import csv

for i in [1, 2, 3]:
    print("helloworld" + str(i))
    result_range = {}
    result_average = {}
    for j in range(1,21):
        filename = "helloworld.json_" + str(i) + "_" + str(j) +".csv"
        maps = {}
        with open(filename) as csvfile:
            csv_reader = csv.reader(csvfile)
            for row in csv_reader:
                if row[0] not in maps:
                    maps[row[0]] = []
                maps[row[0]].append(int(row[1]))
        for key in maps.keys():
            maps[key].sort()
            dis = maps[key][-1] - maps[key][0]
            if key not in result_range:
                 result_range[key] = []
                 result_average[key] = []
            result_range[key].append(dis / maps[key][0])
            result_average[key].append(sum(maps[key]) / len(maps[key]))
    for key in result_range:
        print("The max-min range of function " + key + " is:")
        print(" ".join(str(c) for c in result_range[key]))
        print("The average latency of function " + key + " is ")
        print(" ".join(str(c) for c in result_average[key]))