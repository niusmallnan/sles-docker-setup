# Use SUSE BCI as base for testing our SLES setup tool
FROM registry.suse.com/bci/bci-base:15.7

# Copy our setup binary
COPY bin/setup-docker /usr/local/bin/setup-docker

# Give execute permission
RUN chmod +x /usr/local/bin/setup-docker

# Set working directory
WORKDIR /root

# Default command: just start bash so user can manually test
CMD ["/bin/bash"]
