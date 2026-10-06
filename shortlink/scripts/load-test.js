import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 5 },
        { duration: '1m', target: 10 },
        { duration: '30s', target: 0 },
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
            '请通过 SHORT_CODE 环境变量提供一个真实存在的短码'
        );
    }

    const response = http.get(
        `${BASE_URL}/${SHORT_CODE}`,
        {
            redirects: 0,
            tags: {
                endpoint: 'redirect',
            },
        }
    );

    check(response, {
        '状态码为 301 或 302': (res) =>
            res.status === 301 || res.status === 302,

        '存在 Location 响应头': (res) =>
            Boolean(res.headers.Location),

        '响应耗时小于 1 秒': (res) =>
            res.timings.duration < 1000,
    });

    sleep(0.2);
}