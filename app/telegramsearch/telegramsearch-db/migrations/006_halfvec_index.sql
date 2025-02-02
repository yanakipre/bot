CREATE INDEX embeddings_2000_idx ON embeddings USING hnsw(embedding vector_l2_ops);
