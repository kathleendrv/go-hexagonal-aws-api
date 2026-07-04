import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. Configuración dinámica de escenarios mediante variables de entorno
export const options = {
    stages: [
        { duration: __ENV.STAGE_DURATION || '1m', target: parseInt(__ENV.VIRTUAL_USERS || '10') }
    ],
    thresholds: {
        http_req_failed: ['rate<0.05'], // Alerta si hay más del 5% de errores (Actividad 4)
    },
};

export default function () {
    //  REEMPLAZA ESTA URL POR TU API GATEWAY ACTIVO (Sin la barra / al final)
    const baseUrl = 'https://ffyungh2ue.execute-api.us-east-1.amazonaws.com'; 
    const url = `${baseUrl}/notifications/send`;

    const payload = jsonEncode({
        email: "test@test.com",
        subject: "Prueba de Carga UTESA",
        message: "Mensaje de carga concurrente evaluando escalabilidad Serverless."
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Realizar la petición POST
    const res = http.post(url, payload, params);

    // Validar que AWS responda con HTTP 200
    check(res, {
        'status es 200': (r) => r.status === 200,
    });

    // Pequeño respiro de 100ms entre peticiones por usuario virtual para simular tráfico real
    sleep(0.1); 
}

function jsonEncode(obj) {
    return JSON.stringify(obj);
}