const ctx = document.getElementById('ecoChart').getContext('2d');


const chart = new Chart(ctx, {
    type: 'line',
    data: {
        labels: [],
        datasets: [{
            label: 'Температура (°C)',
            data: [],
            borderColor: '#38bdf8',
            backgroundColor: 'rgba(56, 189, 248, 0.1)',
            borderWidth: 3,
            tension: 0.4, // Плавні лінії
            fill: true,
            pointBackgroundColor: '#38bdf8',
            pointRadius: 4
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { labels: { color: '#f8fafc' } }
        },
        scales: {
            y: {
                grid: { color: 'rgba(255, 255, 255, 0.1)' },
                ticks: { color: '#94a3b8' },
                beginAtZero: false
            },
            x: {
                grid: { display: false },
                ticks: { color: '#94a3b8' }
            }
        }
    }
});

async function fetchData() {
    try {
        const response = await fetch(`http://localhost:8080/measurements?t=${Date.now()}`);
        const data = await response.json();

        if (!data || data.length === 0) return;

        
        const recentData = data.slice(0, 15).reverse();

        
        chart.data.labels = recentData.map(m => new Date(m.created_at || m.CreatedAt).toLocaleTimeString());
        chart.data.datasets[0].data = recentData.map(m => m.value || m.Value);

        chart.update('active'); 
    } catch (error) {
        console.error("API Error:", error);
    }
}


fetchData();
setInterval(fetchData, 5000);

async function verifyAndFillTable(recentData) {
    const tableBody = document.getElementById('tableBody');
    tableBody.innerHTML = ''; 

    const top5 = recentData.slice(-5).reverse();

    for (const m of top5) {
        const row = document.createElement('tr');
        const sensorId = m.sensor_id || m.SensorID;
        
        row.innerHTML = `
            <td>${sensorId}</td>
            <td>${(m.value || m.Value).toFixed(2)}°C</td>
            <td id="status-${sensorId}" class="status-checking">Перевірка...</td>
        `;
        tableBody.appendChild(row);

        try {
            const vRes = await fetch(`http://localhost:8080/measurements/${sensorId}/verify`);
            const vData = await vRes.json();
            const statusCell = document.getElementById(`status-${sensorId}`);
            
            if (vData.is_trusted) {
                statusCell.innerHTML = " Trusted";
                statusCell.className = "status-trusted";
            } else {
                statusCell.innerHTML = " Unverified";
                statusCell.style.color = "#ef4444";
            }
        } catch (e) {
            console.error("Помилка валідації:", e);
        }
    }
}

