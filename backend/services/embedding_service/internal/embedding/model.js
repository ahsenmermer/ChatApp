import pkg from "@xenova/transformers";
const { pipeline } = pkg;

const input = process.argv[2];

async function run() {
  // Doğru pipeline: "embeddings"
  const embed = await pipeline("embeddings", "Xenova/all-MiniLM-L6-v2");
  const output = await embed(input);

  console.log(JSON.stringify({ embedding: output.embedding }));
}

run();
