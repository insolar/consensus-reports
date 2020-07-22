FROM debian:buster-slim
ADD bin/metricreplicator /bin/
ADD bin/report /bin/
RUN chmod +x /bin/metricreplicator /bin/report
