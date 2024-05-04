import http from 'k6/http';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '1m', target: 100 }, // поднимаем нагрузку до 100 в течение 1 минуты
        { duration: '3m', target: 100 }, // поддерживаем нагрузку на уровне 100 в течение 3 минут
        { duration: '1m', target: 0 },   // плавный выход
    ],
};

export default function () {
    let url = 'http://localhost:8000/debug/pprof/';
    let res = http.post(url, JSON.stringify({ /* ваш JSON payload */ }), {
        headers: {
            'Content-Type': 'application/json',
        },
    });

    check(res, {
        'status was 200': (r) => r.status === 200,
    });
}