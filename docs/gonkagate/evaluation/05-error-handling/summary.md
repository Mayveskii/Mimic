=== Error Handling Test ===

--- invalid_key ---
Status: 401, Time: 0.233s
Body: {"error":{"message":"Invalid credentials.","type":"authentication_error","code":"invalid_api_key","param":null}}

--- invalid_model ---
Status: 404, Time: 0.281s
Body: {"error":{"message":"Model not available.","type":"invalid_request_error","code":"model_not_found","param":null,"metadata":{"model_slug":"nonexistent/model"}}}

--- timeout ---
Status: URL_ERROR, Time: 0.073s
Body: timed out

--- rate_limit_burst (10 requests) ---
Non-200 responses: 0
Statuses: [200, 200, 200, 200, 200, 200, 200, 200, 200, 200]
