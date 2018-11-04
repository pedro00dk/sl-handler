FROM node:8-alpine

# Create app directory
WORKDIR /app

RUN npm install -g express

# Bundle app source
COPY . .

RUN npm install --only=production
RUN npm link express

#Your app binds to port 8080 so you'll use the EXPOSE instruction to have it mapped by the docker daemon:
EXPOSE 8080

CMD [ "node", "./server.js" ]
