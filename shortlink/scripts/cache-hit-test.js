import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const rateLimited = new Counter('rate_limited_requests');
const unexpectedStatus = new Counter('unexpected_status_requests');

export const options = {
    stages: [
        { duration: '15s', target: 1 },
        { duration: '30s', target: 5 },
        { duration: '15s', target: 0 },
    ],

    thresholds: {
        http_req_duration: ['p(95)<1000'],
        checks: ['rate>0.95'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SHORT_CODE = __ENV.SHORT_CODE;

export default function () {
    if (!SHORT_CODE) {
        throw new Error(
            '请设置 SHORT_CODE 环境变量，使用真实存在的短码'
        );
    }

    const response = http.get(
        `${BASE_URL}/${SHORT_CODE}`,
        {
            redirects: 0,
            tags: {
                endpoint: 'redirect',
                scenario: 'cache-hit',
            },
        }
    );

    if (response.status === 429) {
        rateLimited.add(1);
    }

    if (response.status < 300 || response.status >= 400) {
        unexpectedStatus.add(1);
    }

    check(response, {
        '返回 3xx 重定向状态': (res) =>
            res.status >= 300 && res.status < 400,

        '包含 Location 响应头': (res) =>
            Boolean(res.headers.Location),

        '请求耗时小于 1 秒': (res) =>
            res.timings.duration < 1000,
    });

    sleep(0.2);
}