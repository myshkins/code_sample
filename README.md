## Overview
This is a program I wrote recently for a job application coding challenge. It is a very basic endpoint monitoring program, designed to monitor the availability of various http endpoints passed in by the user.

You can pass a filepath argument that points to a yaml containing the enpoints you would like to monitor, or simply use the default sample yaml provided.

### How to Run
Copy the source code to your local machine
```
git clone git@github.com:myshkins/code_sample.git

cd code_sample
```
If you have go installed on your machine, you can follow the steps in the next section. Alternatively, you may follow the steps in [With Docker](###with-docker).

##### With Go installed
Change to the `cmd/health_check` directory and then build the binary
```
cd cmd/health_check

go build -C cmd/health_check
```

You can then run the program via:
```
./health_check --config-file=path/to/config/yaml --interval=15
```
Both flags are optional. The `-config-file` flag defaults to the `input.yaml` provided in the repo, and the `-interval` flag defaults to 15 seconds.

##### With Docker
To run the program with docker, first build the image. From the root of the `fetch_takehome`  repo, run:
```
docker build -t health_check
```

You can then run the program in a docker container via:
```
docker run -v "path/to/input.yaml:input.yaml:ro" health_check "--interval=15"
```
Again both flags are optional.

