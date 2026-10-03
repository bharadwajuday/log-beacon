# Future Enhancements for Log Beacon

This file tracks potential next steps and features to improve the Log Beacon platform.

1. - [x] **Refine the Search Query**:
   - Supported structured queries like `level:error AND service:api-gateway`.
   - **Status**: Completed. Implemented `AND` operator, automatic label rewriting, and frontend integration.

2. - [x] **Implement Log Retention & Compaction in Hot Storage**:
   - Added periodic background retention worker to `hot-storage` (`PurgeExpiredLogs`) scanning BadgerDB and purging expired documents from both Bleve index and Badger storage.
   - Configurable retention periods via `HOT_STORAGE_RETENTION` (e.g. 12h, 24h) and purge frequencies via `HOT_STORAGE_PURGE_INTERVAL`.
   - **Status**: Completed.

3. - [x] **Build a "Live Tail" Feature**:
   - Added a WebSocket endpoint to the `api` service (`/api/v1/tail`) subscribing to the NATS JetStream log stream.
   - Integrated live tail toggle directly into the React UI with real-time streaming.
   - **Status**: Completed.

4. - [x] **Cold Storage Querying & Columnar Analytics**:
   - Replaced single-log `.gz` writes with in-memory dual-trigger batching and Apache Parquet columnar serialization.
   - Built cold query engine (`internal/coldquery`) scanning partitioned Parquet files in MinIO with projection and predicate filtering.
   - Wired federated search routing directly into `/api/v1/search` supporting seamless queries across hot and cold archives.
   - **Status**: Completed.

5. - [x] **Implement Authentication & Authorization**:
   - **Backend**: Implemented JWT authentication, password hashing with bcrypt, Postgres user repository, and auth middleware on `/search` and `/tail`.
   - **Frontend**: Added login & registration UI modal, token persistence in `localStorage`, and authenticated HTTP/WebSocket clients.
   - **Status**: Completed. Future work: API Key authentication for the `/api/v1/ingest` endpoint.

---

## UI Improvements

1.  **Improved Search Experience:**
    -   **Search History:** Store recent searches in `localStorage` and display them as suggestions.
    -   **Debounced Search:** Automatically trigger the search as the user types, with a debounce to prevent excessive API calls.
    -   **Date/Time Range Picker:** Add a date picker to filter logs within a specific time window.

2.  **Enhanced Results Display:**
    -   **Click-to-Expand Rows:** Allow users to click a log row to expand it and view detailed information, such as all labels.
    -   **Highlighting Search Terms:** Highlight the matching query terms within the displayed log messages.

3.  **Real-time Features:**
    -   [x] **"Live Tail" Toggle:** Add a UI switch to connect to a WebSocket and stream logs in real-time.

4.  **Usability and Polish:**
    -   **Clear Search Button:** Add an "X" icon to the search bar to clear the input.
    -   **Favicon:** Add a custom favicon for the browser tab.
