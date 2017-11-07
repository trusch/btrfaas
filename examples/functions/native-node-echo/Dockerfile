FROM node

COPY package.json /app/
WORKDIR /app
RUN npm install

COPY server.js /app/
ADD https://raw.githubusercontent.com/trusch/btrfaas/master/frunner/grpc/frunner.proto /app/

CMD ["node","server.js"]
