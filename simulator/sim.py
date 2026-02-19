import requests
import time
import random


API_URL = "http://api:8080/measurements"
SENSORS = ["sensor_kiev", "sensor_lviv", "sensor_odesa", "sensor_kharkiv"]

print("Симулятор датчиків запущено. Очікування старту API (10 секунд)...")
time.sleep(10)

while True:
   
    data = {
        "sensor_id": random.choice(SENSORS),
        "value": round(random.uniform(-5.0, 35.5), 2)
    }
    
    try:
        response = requests.post(API_URL, json=data)
        print(f"Відправлено: {data} | Статус: {response.status_code}")
    except Exception as e:
        print(f"Помилка з'єднання: {e}")
    
    time.sleep(13)