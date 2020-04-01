FROM eciavatta/caronte-env

COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

CMD ./caronte
