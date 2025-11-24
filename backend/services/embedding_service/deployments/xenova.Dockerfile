FROM node:18
WORKDIR /app/internal/embedding
COPY internal/embedding/package*.json ./
RUN npm ci --production
COPY internal/embedding/*.js ./
EXPOSE 3000
CMD ["node", "server.js"]
