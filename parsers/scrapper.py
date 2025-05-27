import os
import json
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException, WebDriverException
from webdriver_manager.chrome import ChromeDriverManager

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
INPUT_FILE = os.path.join(BASE_DIR,"input.json")
OUTPUT_FILE = os.path.join(BASE_DIR,"output.json")

if not os.path.exists(INPUT_FILE):
    print(f"File {INPUT_FILE} was not founded.")
    exit(1)

with open(INPUT_FILE, "r", encoding="utf-8") as f:
    try:
        urls = json.load(f)
        if not isinstance(urls, list):
            raise ValueError("JSON has to include list of URL.")
    except Exception as e:
        print(f"Error reading JSON: {e}")
        exit(1)

options = Options()
options.add_argument("--headless")
options.add_argument("--disable-gpu")
options.add_argument("--no-sandbox")
options.add_argument("--disable-dev-shm-usage")
options.add_argument("--disable-blink-features=AutomationControlled")

results = {}

try:
    driver = webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=options)

    for url in urls:
        try:
            driver.get(url)
            wait = WebDriverWait(driver, 30)
            description_elem = wait.until(
                EC.presence_of_element_located((By.CLASS_NAME, "job-post__description"))
            )
            results[url] = description_elem.text.strip()
        except Exception:
            results[url] = None

    driver.quit()

except WebDriverException as e:
    print(f"Error starting WebDriver: {e}")
    exit(1)

with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
    json.dump(results, f, ensure_ascii=False, indent=2)

print(f"Done. Data was saved in {OUTPUT_FILE}")
