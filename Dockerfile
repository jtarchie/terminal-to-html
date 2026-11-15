FROM public.ecr.aws/docker/library/golang:1.26.10@sha256:672c944cfd07599b5f1bb179869270cdc94c6254e41db4d431259693cb286b82

RUN apt-get update -q && apt-get install -y zip

WORKDIR /src
ADD . /src

CMD [ "make", "dist" ]
