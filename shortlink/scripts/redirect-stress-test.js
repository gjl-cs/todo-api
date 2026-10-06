import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const redirectSuccesses = new Counter('redirect_successes');
const rateLimitedRequests = new Counter('rate_limited_requests');
const unexpectedStatusRequests = new Counter('unexpected_status_requests');

const VUS = Number(__ENV.VUS || 100);
const DURATION = __ENV.DURATION || '30s';

export const options = {
  vus: VUS,
  duration: DURATION,
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    unexpected_status_requests: ['count==0'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SHORT_CODE = __ENV.SHORT_CODE;

export default function () {
  if (!SHORT_CODE) {
    throw new Error('请设置 SHORT_CODE 环境变量');
  }

  const response = http.get(
    `${BASE_URL}/${SHORT_CODE}`,
    {
      redirects: 0,
      tags: { endpoint: 'redirect' },
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

    console.log(
      `UNEXPECTED: status=${response.status}, ` +
      `error=${response.error || 'none'}, ` +
      `error_code=${response.error_code || 'none'}, ` +
      `duration=${response.timings.duration.toFixed(2)}ms`
    );
  }

  check(response, {
    '返回正常重定向或明确限流': () =>
      isRedirect || isRateLimited,

    '正常重定向包含 Location': () =>
      !isRedirect || Boolean(response.headers.Location),

    '响应耗时小于 1 秒': (res) =>
      res.timings.duration < 1000,
  });
}