FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm install
CMD ["node", "packages/backend/dist/server.js"]
