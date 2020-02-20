FROM debian:jessie

RUN apt-get update && apt-get -y install cron
WORKDIR /sync
COPY disableTool /sync/
COPY conf.json /sync/
COPY checkCronList /sync/
RUN mkdir -p /sync/log
EXPOSE 6060
ENTRYPOINT ["/sync/checkCronList","-mysqlDomain","-username","-password","-apiserverDomain"]

