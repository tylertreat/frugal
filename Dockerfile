FROM drydock-prod.workiva.net/workiva/messaging-docker-images:210905 as build

ARG GIT_SSH_KEY
ARG KNOWN_HOSTS_CONTENT
WORKDIR /go/src/github.com/Workiva/frugal/
ADD . /go/src/github.com/Workiva/frugal/

RUN mkdir /root/.ssh && \
    echo "$KNOWN_HOSTS_CONTENT" > "/root/.ssh/known_hosts" && \
    chmod 700 /root/.ssh/ && \
    umask 0077 && echo "$GIT_SSH_KEY" >/root/.ssh/id_rsa && \
    eval "$(ssh-agent -s)" && ssh-add /root/.ssh/id_rsa

ARG GOPATH=/go/
ENV PATH $GOPATH/bin:$PATH
RUN git config --global url.git@github.com:.insteadOf https://github.com
ENV FRUGAL_HOME=/go/src/github.com/Workiva/frugal
ENV CODECOV_TOKEN='d88d0bbe-b5f0-4dce-92ae-a110aa028ddb'
RUN echo "Starting the script section" && \
		./scripts/smithy.sh && \
		cat $FRUGAL_HOME/test_results/smithy_dart.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_go.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_generator.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_python.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_java.sh_out.txt && \
		echo "script section completed"

ARG BUILD_ARTIFACTS_RELEASE=/go/src/github.com/Workiva/frugal/frugal
ARG BUILD_ARTIFACTS_BUILD=/go/src/github.com/Workiva/frugal/python2_pip_deps.txt:/go/src/github.com/Workiva/frugal/python3_pip_deps.txt
ARG BUILD_ARTIFACTS_GO_LIBRARY=/go/src/github.com/Workiva/frugal/goLib.tar.gz
ARG BUILD_ARTIFACTS_PYPI=/go/src/github.com/Workiva/frugal/frugal-*.tar.gz
ARG BUILD_ARTIFACTS_ARTIFACTORY=/go/src/github.com/Workiva/frugal/frugal-*.jar
ARG BUILD_ARTIFACTS_PUB=/go/src/github.com/Workiva/frugal/frugal.pub.tgz
ARG BUILD_ARTIFACTS_TEST_RESULTS=/go/src/github.com/Workiva/frugal/test_results/*

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/frugal .

FROM scratch
COPY --from=build /bin/frugal /bin/frugal
ENTRYPOINT ["frugal"]
