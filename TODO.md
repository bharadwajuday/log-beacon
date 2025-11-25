# Future Enhancements for Log Beacon

This file tracks potential next steps and features to improve the Log Beacon platform.

1.- [x] **Refine the Search Query**:
  - Currently, the search is a simple query string.
  - We need to support structured queries like `level:error AND service:api-gateway`.
  - This will require parsing the query string in the backend and constructing a more complex Bleve query.
  - **Status**: Completed. Implemented `AND` operator, automatic label rewriting, and frontend integration.ies in the `hot-storage` service.

2.  **Implement Log Retention in Hot Storage:**
    -   Add a mechanism to the `hot-storage` service to periodically purge old data from Bleve and BadgerDB.
    -   This will keep the hot storage index lean, fast, and prevent it from growing indefinitely.
    -   A time-based retention policy (e.g., keep last 24 hours) is a good starting point.

3.  **Build a "Live Tail" Feature:**
    -   [x] Add a WebSocket endpoint to the `api` service.
    -   [x] This endpoint would subscribe to the NATS stream and stream logs to a connected client in real-time, providing a `tail -f` like experience.
    -   **Status**: Completed. Implemented WebSocket endpoint at `/api/v1/tail`.

4.  **Explore the Cold Storage:**
    -   Build a mechanism to search the "cold" data stored in MinIO.
    -   This would likely be a slower, asynchronous process initiated via a separate API endpoint (e.g., `/api/v1/archive/search`).
    -   The process would involve downloading, decompressing, and searching through the gzipped log chunks in the MinIO bucket.

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
