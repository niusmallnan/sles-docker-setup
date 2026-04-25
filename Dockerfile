# Use SUSE BCI as base for testing our SLES setup tool
FROM docker:cli

# Copy our setup binary
COPY bin/docker-pilot /usr/local/bin/docker-pilot

# Give execute permission
RUN chmod +x /usr/local/bin/docker-pilot

# Set working directory
WORKDIR /root

