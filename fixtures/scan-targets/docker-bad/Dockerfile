FROM ubuntu:22.04

USER root

ENV AWS_ACCESS_KEY_ID=not-a-real-access-key-for-sarif-fixtures
ENV AWS_SECRET_ACCESS_KEY=not-a-real-secret-for-sarif-fixtures

ADD https://example.com/install.sh /tmp/install.sh

RUN apt-get update && apt-get install -y curl openssh-server

EXPOSE 22

CMD ["sh", "-c", "curl http://example.com/bootstrap.sh | sh"]
