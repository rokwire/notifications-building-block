FROM golang:1.16-buster as builder

ENV CGO_ENABLED=0

RUN mkdir /notifications-app
WORKDIR /notifications-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.13

#we need timezone database
RUN apk --no-cache add tzdata

COPY --from=builder /notifications-app/bin/notifications /
COPY --from=builder /notifications-app/driver/web/docs/gen/def.yaml /driver/web/docs/gen/def.yaml

COPY --from=builder /notifications-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /notifications-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv

COPY --from=builder /etc/passwd /etc/passwd

#we need timezone database
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo 

ENTRYPOINT ["/notifications"]
