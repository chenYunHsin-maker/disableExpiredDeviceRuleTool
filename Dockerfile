FROM golang:1.13.5
WORKDIR /sync
COPY disableTool.go /sync/
COPY checkCronList.go /sync/
RUN go get -u github.com/go-sql-driver/mysql
RUN go get -u github.com/syhlion/sqlwrapper
RUN go get -u github.com/golang/glog
RUN cd /sync && go build disableTool.go
RUN go build checkCronList && mkdir -p log
EXPOSE 6060
ENTRYPOINT ./checkCronList
#ENTRYPOINT  ./sync -log_dir=/sync/log -alsologtostderr 127.0.0.1:3308 root root http://127.0.0.1:8080 2020-01-01 2020-12-31
#CMD ["cd /sync", "mkdir -p log" ,"go run disableTool.go -log_dir=log -alsologtostderr","1","2","3","4"]

