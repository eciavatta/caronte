FROM eciavatta/caronte-env:latest

ENV GIN_MODE release

COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

RUN npm i --global yarn && cd frontend && yarn install && yarn build

CMD ./caronte
