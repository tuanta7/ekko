FROM ubuntu:latest
LABEL authors="tuanta"

ENTRYPOINT ["top", "-b"]