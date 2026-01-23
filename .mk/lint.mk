lint:: | mise_install
	 if [ -d provider ]; then \
		pushd provider && golangci-lint run --timeout 10m && popd ; \
	 fi
	 if [ -d examples ]; then \
		pushd examples && golangci-lint run --timeout 10m --build-tags all && popd ; \
	fi
