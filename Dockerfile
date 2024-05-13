FROM golang:1.20-alpine

# Set maintainer label:
LABEL maintainer='S2310455017@fhooe.at'

# Set working directory: `/src`
WORKDIR /src

# Copy local files to the working directory
COPY app.go go.* ./
COPY cart.go ./
COPY user.go ./
COPY model.go ./
COPY main.go schema.sql ./


# Build the GO app as myapp binary and move it to /usr/
RUN CGO_ENABLED=0 go build -o /usr/microS

#Expose port
EXPOSE 5416

# Run the service myapp when a container of this image is launched
CMD ["/usr/microS"]