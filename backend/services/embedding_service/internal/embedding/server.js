import express from "express";
import bodyParser from "body-parser";
import { pipeline, env } from "@xenova/transformers";

// Model cache ayarları
env.allowLocalModels = false;
env.useBrowserCache = false;

const PORT = process.env.PORT || 3000;
const app = express();
app.use(bodyParser.json({ limit: "10mb" }));

let embedModel = null;
let modelLoading = false;

async function loadModel() {
  if (modelLoading) {
    console.log("Model already loading...");
    return;
  }

  modelLoading = true;

  try {
    console.log("📦 Loading Xenova embedding model...");
    console.log("📦 Model: Xenova/all-MiniLM-L6-v2");

    // "feature-extraction" kullan - bu kesin çalışır
    embedModel = await pipeline("embeddings", "Xenova/all-MiniLM-L6-v2");

    console.log("✅ Model loaded successfully!");
    modelLoading = false;
    return embedModel;
  } catch (error) {
    console.error("❌ Model loading failed:", error.message);
    console.error("Stack:", error.stack);
    modelLoading = false;
    throw error;
  }
}

app.post("/embed", async (req, res) => {
  try {
    const { text } = req.body;

    if (!text) {
      return res.status(400).json({ error: "text field is required" });
    }

    if (!embedModel) {
      return res.status(503).json({
        error: "model is still loading",
        loading: modelLoading,
      });
    }

    // Embedding oluştur
    const output = await embedModel(text, {
      pooling: "mean",
      normalize: true,
    });

    // Output her zaman tensor formatında gelir
    // .data ile raw array'e çevir
    const vector = Array.from(output.data);

    console.log(`📊 Generated embedding with dimension: ${vector.length}`);

    res.json({
      embedding: vector,
      dimension: vector.length,
    });
  } catch (err) {
    console.error("❌ Embedding error:", err);
    res.status(500).json({
      error: err.message || "internal server error",
    });
  }
});

app.get("/health", (req, res) => {
  if (embedModel) {
    res.json({ status: "ready", model: "Xenova/all-MiniLM-L6-v2" });
  } else if (modelLoading) {
    res.json({ status: "loading", model: "Xenova/all-MiniLM-L6-v2" });
  } else {
    res.status(503).json({ status: "not_ready" });
  }
});

// Test endpoint
app.get("/test", async (req, res) => {
  try {
    if (!embedModel) {
      return res.status(503).json({ error: "model not ready" });
    }

    const testText = "Hello world";
    const output = await embedModel(testText, {
      pooling: "mean",
      normalize: true,
    });
    const vector = Array.from(output.data);

    res.json({
      test: "success",
      text: testText,
      dimension: vector.length,
      sample: vector.slice(0, 5), // İlk 5 değeri göster
    });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Server'ı başlat
async function startServer() {
  try {
    await loadModel();

    app.listen(PORT, "0.0.0.0", () => {
      console.log(`🚀 Xenova embedding server running on port ${PORT}`);
      console.log(`📍 Health check: http://localhost:${PORT}/health`);
      console.log(`📍 Test endpoint: http://localhost:${PORT}/test`);
    });
  } catch (error) {
    console.error("💥 Failed to start server:", error);
    process.exit(1);
  }
}

startServer();
