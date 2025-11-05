# Future Enhancements for Log Beacon

This file tracks potential next steps and features to improve the Log Beacon platform.

1.  **Refine the Search Query:**
    -   Enhance the search API (`/api/v1/search`) to support structured queries on labels (e.g., `level:error AND service:api-gateway`).
    -   This will involve parsing the query string and building more complex Bleve queries in the `hot-storage` service.

2.  **Implement Log Retention in Hot Storage:**
    -   Add a mechanism to the `hot-storage` service to periodically purge old data from Bleve and BadgerDB.
    -   This will keep the hot storage index lean, fast, and prevent it from growing indefinitely.
    -   A time-based retention policy (e.g., keep last 24 hours) is a good starting point.

3.  **Build a "Live Tail" Feature:**
    -   Add a WebSocket endpoint to the `api` service.
    -   This endpoint would subscribe to the NATS stream and stream logs to a connected client in real-time, providing a `tail -f` like experience.

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
    -   **"Live Tail" Toggle:** Add a UI switch to connect to a WebSocket and stream logs in real-time.

4.  **Usability and Polish:**
    -   **Clear Search Button:** Add an "X" icon to the search bar to clear the input.
    -   **Favicon:** Add a custom favicon for the browser tab.