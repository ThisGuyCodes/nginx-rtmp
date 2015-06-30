FROM ubuntu

RUN apt-get update ;\
	apt-get install -y \
	build-essential \
	libpcre3 \
	libpcre3-dev \
	libssl-dev

WORKDIR /root

# Downloaded from http://nginx.org/download/nginx-1.8.0.tar.gz
COPY nginx-1.8.0.tar.gz /root/
RUN tar -xvz < nginx-1.8.0.tar.gz

# Downloaded from https://github.com/arut/nginx-rtmp-module/archive/v1.1.7.tar.gz
COPY nginx-rtmp-module-v1.1.7.tar.gz /root/
RUN tar -xvz < nginx-rtmp-module-v1.1.7.tar.gz

WORKDIR /root/nginx-1.8.0

RUN ./configure --with-http_ssl_module --add-module=../nginx-rtmp-module-1.1.7
RUN make
RUN make install

WORKDIR /root

RUN mkdir -p /var/log/nginx

COPY nginx.conf /etc/nginx/nginx.conf

CMD ["/usr/local/nginx/sbin/nginx", "-c", "/etc/nginx/nginx.conf"]

EXPOSE  1935
