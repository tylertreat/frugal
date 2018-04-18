
.PHONY: all
all: install

# Installs dependencies while sym linking the local frugal into the
# vendor folder so local changes can be easily tested
.PHONY: install
install:
	rm -rf vendor && \
		glide install && \
		rm -rf vendor/github.com/Workiva/frugal/lib/go && \
		ln -s ${GOPATH}/src/github.com/Workiva/frugal/lib/go vendor/github.com/Workiva/frugal/lib/go

