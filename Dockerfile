FROM drydock-prod.workiva.net/workiva/messaging-docker-images:210905 as build

ARG PIP_INDEX_URL
ARG PIP_EXTRA_INDEX_URL=https://pypi.python.org/simple/

# Build Environment Vars
ARG BUILD_ID
ARG BUILD_NUMBER
ARG BUILD_URL
ARG GIT_COMMIT
ARG GIT_BRANCH
ARG GIT_TAG
ARG GIT_COMMIT_RANGE
ARG GIT_HEAD_URL
ARG GIT_MERGE_HEAD
ARG GIT_MERGE_BRANCH
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
RUN if [ -d /usr/local/go_appengine ]; then ENV PATH $PATH:/usr/local/go_appengine; fi
ADD ./settings.xml /root/.m2/settings.xml
ENV FRUGAL_HOME=/smithy-builder/builds/Workiva/frugal/cache/GO/src/github.com/Workiva/frugal
ENV SMITHY_ROOT=/smithy-builder/builds/Workiva/frugal/workspace
ENV CODECOV_TOKEN='d88d0bbe-b5f0-4dce-92ae-a110aa028ddb'
RUN echo "Starting the script section" && \
		./scripts/smithy.sh && \
		cat $SMITHY_ROOT/test_results/smithy_dart.sh_out.txt && \
		cat $SMITHY_ROOT/test_results/smithy_go.sh_out.txt && \
		cat $SMITHY_ROOT/test_results/smithy_generator.sh_out.txt && \
		cat $SMITHY_ROOT/test_results/smithy_python.sh_out.txt && \
		cat $SMITHY_ROOT/test_results/smithy_java.sh_out.txt && \
		echo "script section completed"
RUN go buildARG BUILD_ARTIFACTS_RELEASE=/frugal
ARG BUILD_ARTIFACTS_BUILD=/go/src/github.com/Workiva/frugal/python2_pip_deps.txt:/go/src/github.com/Workiva/frugal/python3_pip_deps.txt
ARG BUILD_ARTIFACTS_GO_LIBRARY=/go/src/github.com/Workiva/frugal/goLib.tar.gz
ARG BUILD_ARTIFACTS_PYPI=/go/src/github.com/Workiva/frugal/frugal-*.tar.gz
ARG BUILD_ARTIFACTS_ARTIFACTORY=/go/src/github.com/Workiva/frugal/frugal-*.jar
ARG BUILD_ARTIFACTS_PUB=/go/src/github.com/Workiva/frugal/frugal.pub.tgz
ARG BUILD_ARTIFACTS_TEST_RESULTS=/go/src/github.com/Workiva/frugal/test_results/*

RUN mkdir /audit/
ARG BUILD_ARTIFACTS_AUDIT=/audit/*

RUN pip freeze > /audit/pip.lock
FROM scratch
