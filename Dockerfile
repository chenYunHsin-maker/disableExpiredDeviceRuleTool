FROM debian:jessie

RUN apt-get update && apt-get -y install cron
WORKDIR /sync
COPY disableTool /sync/
COPY checkCronList /sync/
RUN mkdir -p /sync/log
EXPOSE 6060
ENTRYPOINT ["/sync/checkCronList","-mysqlDomain=sdwan-orch-db-orchestrator-db:3306","-username=root","-password=root","-apiserverDomain=http://sdwan-api-01-apiserver:80"]

