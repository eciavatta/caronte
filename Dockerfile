FROM eciavatta/caronte-env:latest

COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

CMD ./caronte
