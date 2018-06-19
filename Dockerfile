FROM golang:1.10-stretch
WORKDIR /go/src/github.com/Workiva/frugal

# One layer caches the deps
RUN curl https://glide.sh/get | sh

ARG GIT_SSH_KEY		      
ARG KNOWN_HOSTS_CONTENT	      
RUN mkdir /root/.ssh && \     
  echo "$GIT_SSH_KEY" > /root/.ssh/id_rsa && \
  echo "$KNOWN_HOSTS_CONTENT" > /root/.ssh/known_hosts && \
  chmod -R 700 /root/.ssh && \
  git config --global url."git@github.com:".insteadOf "https://github.com/"
COPY glide.yaml .
COPY glide.lock .
RUN glide install

# One layer builds the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/frugal .

COPY glide.lock /deps/
ARG BUILD_ARTIFACTS_DEPENDENCIES=/deps/glide.lock

FROM scratch
COPY --from=0 /bin/frugal /bin/frugal
ENTRYPOINT ["frugal"]
