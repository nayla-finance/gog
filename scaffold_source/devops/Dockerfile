FROM cgr.dev/chainguard/busybox:latest

ARG TARGETARCH
WORKDIR /app

# Debug: Show what files are available and what TARGETARCH is
RUN echo "TARGETARCH is: ${TARGETARCH}" && \
    echo "Available files in build context:" && \
    ls -la /

# Copy the pre-built binary for the target architecture
# and make it executable
COPY --chmod=755 service-${TARGETARCH} ./service
COPY docs ./docs
COPY migrations ./migrations

# Debug: Verify the binary was copied and show its permissions
RUN echo "Files in /app after copy:" && \
    ls -la /app && \
    echo "Binary details:" && \
    file /app/service || echo "file command failed"

EXPOSE 3000

ENTRYPOINT ["./service"]
