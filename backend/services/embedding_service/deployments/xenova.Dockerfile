FROM node:18-slim

WORKDIR /app

# Install wget for healthcheck
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Copy package files
COPY internal/embedding/package.json internal/embedding/package-lock.json ./

# Install dependencies
RUN npm ci --omit=dev

# Copy JavaScript files
COPY internal/embedding/server.js ./

ENV PORT=3000

EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=5s --start-period=60s --retries=3 \
    CMD wget -qO- http://localhost:3000/health || exit 1

CMD ["node", "server.js"]