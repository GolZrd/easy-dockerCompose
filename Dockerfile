FROM golang:latest

COPY ./ ./

RUN go mod download
RUN go build -o main .

CMD [ "./main" ]