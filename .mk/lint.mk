lint:: | mise_install
	 if [ -d provider ]; then \
		(cd provider && golangci-lint run --timeout 10m); \
	 fi
	 if [ -d examples ]; then \
		(cd examples && golangci-lint run --timeout 10m --build-tags all); \
	fi
