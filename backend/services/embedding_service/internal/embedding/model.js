import { pipeline } from "@xenova/transformers";

async function generateEmbedding() {
  try {
    // stdin'den text oku
    const input = process.argv[2];

    if (!input || input.trim() === "") {
      console.error("Error: No input text provided");
      process.exit(1);
    }

    // Embedding pipeline oluştur
    const extractor = await pipeline(
      "feature-extraction",
      "Xenova/all-MiniLM-L6-v2"
    );

    // Embedding oluştur
    const output = await extractor(input, { pooling: "mean", normalize: true });

    // Tensor'u array'e çevir
    const embeddingArray = Array.from(output.data);

    // JSON olarak yazdır
    console.log(
      JSON.stringify({
        embedding: embeddingArray,
        dimension: embeddingArray.length,
      })
    );

    process.exit(0);
  } catch (error) {
    console.error("Error generating embedding:", error.message);
    process.exit(1);
  }
}

generateEmbedding();
