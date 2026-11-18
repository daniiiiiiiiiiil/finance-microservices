import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8081";
const EMAIL = __ENV.TEST_EMAIL;
const PASSWORD = __ENV.TEST_PASSWORD;

const failedRequests = new Rate("failed_requests");
const apiLatency = new Trend("api_latency", true);

export const options = {
    stages: [
        { duration: "30s", target: 25 },
        { duration: "30s", target: 50 },
        { duration: "30s", target: 75 },
        { duration: "30s", target: 100 },
        { duration: "60s", target: 100 },
        { duration: "30s", target: 0 },
    ],
    thresholds: {
        http_req_failed: ["rate<0.02"],
        http_req_duration: ["p(95)<1500", "p(99)<3000"],
        failed_requests: ["rate<0.02"],
    },
};

export function setup() {
    if (!EMAIL || !PASSWORD) {
        throw new Error("Set TEST_EMAIL and TEST_PASSWORD");
    }

    const response = http.post(
        `${BASE_URL}/api/v1/auth/login`,
        JSON.stringify({ email: EMAIL, password: PASSWORD }),
        { headers: { "Content-Type": "application/json" } }
    );

    check(response, { "login successful": r => r.status === 200 });

    if (response.status !== 200) {
        throw new Error(`Login failed: ${response.status} ${response.body}`);
    }

    const body = response.json();
    const token = body.token || body.data?.token;

    if (!token) {
        throw new Error("JWT token not found in response");
    }

    return { token };
}

export default function (data) {
    const endpoints = [
        "/api/v1/finance/dashboard",
        "/api/v1/finance/analytics?period=month",
        "/api/v1/finance/categories",
        "/api/v1/finance/transactions?limit=20&offset=0",
        "/api/v1/shopping?limit=20&offset=0",
    ];

    const path = endpoints[(__VU + __ITER) % endpoints.length];

    const response = http.get(`${BASE_URL}${path}`, {
        headers: {
            Authorization: `Bearer ${data.token}`,
        },
    });

    const success = check(response, {
        "HTTP 200": r => r.status === 200,
    });

    failedRequests.add(!success);
    apiLatency.add(response.timings.duration);

    sleep(0.5);
}