import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

// 自定义统计指标
const redirectSuccesses = new Counter('redirect_successes');
const rateLimitedRequests = new Counter('rate_limited_requests');
const unexpectedStatusRequests = new Counter('unexpected_status_requests');

// 从环境变量读取并发数和测试时长
const VUS = Number(__ENV.VUS || 20);
const DURATION = __ENV.DURATION || '30s';

export const options = {
    vus: VUS,
    duration: DURATION,

    thresholds: {
        http_req_duration: ['p(95)<1000'],
        checks: ['rate>0.99'],
        unexpected_status_requests: ['count==0'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SHORT_CODE = __ENV.SHORT_CODE;

export default function () {
    if (!SHORT_CODE) {
        throw new Error(
            '请设置 SHORT_CODE 环境变量，填写一个真实存在的短链接短码'
        );
    }

    const response = http.get(
        `${BASE_URL}/${SHORT_CODE}`,
        {
            // 不自动跟随重定向，否则测到的是目标网站的响应
            redirects: 0,
            tags: {
                endpoint: 'redirect',
            },
        }
    );

    const isRedirect =
        response.status >= 300 &&
        response.status < 400 &&
        Boolean(response.headers.Location);

    const isRateLimited = response.status === 429;

    if (isRedirect) {
        redirectSuccesses.add(1);
    } else if (isRateLimited) {
        rateLimitedRequests.add(1);
    } else {
        unexpectedStatusRequests.add(1);
    }

    check(response, {
        '返回正常重定向或明确限流': () =>
            isRedirect || isRateLimited,

        '正常重定向包含 Location': () =>
            !isRedirect || Boolean(response.headers.Location),

        '响应耗时小于 1 秒': (res) =>
            res.timings.duration < 1000,
    });

    sleep(0.2);
}