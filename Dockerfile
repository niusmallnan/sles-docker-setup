# Use SUSE BCI as base for testing our SLES setup tool
FROM registry.suse.com/bci/bci-base:15.7

# Copy our setup binary
COPY bin/docker-pilot /usr/local/bin/docker-pilot

# Give execute permission
RUN chmod +x /usr/local/bin/docker-pilot

# Set working directory
WORKDIR /root

# Default command: just start bash so user can manually test
CMD ["/bin/bash"]
