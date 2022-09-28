FROM ubuntu:16.04

RUN apt update
RUN apt upgrade -y
RUN apt-get install build-essential -y

# Install go
RUN apt install wget -y
RUN wget https://dl.google.com/go/go1.18.6.linux-amd64.tar.gz
RUN tar -xvf go1.18.6.linux-amd64.tar.gz 
RUN mv go /usr/local 

ENV GOROOT=/usr/local/go 
ENV PATH=$GOROOT/bin:$PATH

# Install IPFS
RUN wget https://dist.ipfs.io/go-ipfs/v0.14.0/go-ipfs_v0.14.0_linux-amd64.tar.gz
RUN tar -xzf go-ipfs_v0.14.0_linux-amd64.tar.gz
RUN ./go-ipfs/ipfs init

WORKDIR /app/

COPY . .

EXPOSE 8001

ENTRYPOINT ["tail", "-f", "/dev/null"]