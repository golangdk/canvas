create index on newsletters using gin
  (to_tsvector('english', title || ' ' || body));
