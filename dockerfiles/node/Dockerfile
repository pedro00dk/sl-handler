FROM node:8-alpine

# Create app directory
WORKDIR /app

RUN npm install -g express@4.16.1 && \
    npm link express

# Bundle app source
COPY . .

RUN npm install --only=production

#Your app binds to port 8080 so you'll use the EXPOSE instruction to have it mapped by the docker daemon:
EXPOSE 8080

CMD [ "node", "server.js" ]

