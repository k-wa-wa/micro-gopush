
curl -d '{"subscription":{"endpoint":"https://fcm.googleapis.com/fcm/send/e3ADCkZfb6g:APA91bEM7l81tERM3VNKh9g9immMYWo_vYzg6V-c18necq2OL3jAqOs-Favty5Z6LQrGcBBhAR4MLaeDF8qbxRNrM1RiK3C6V_8scH-aX_hMHzWX_d9NxV452vJQ0ZHnHNN2nBhGUNql","expirationTime":null,"keys":{"p256dh":"BApDBHfM8X7vMjrO7Fc9ZaODaseHT2tEJXZDDg_65K1AbNHXJj8biZ-BHHkvyVpaWm6037JkvEIupkDQHCL1r8g","auth":"MDUdFwPhU1IjcpyO3XL6sw"}}}' \
    -XPOST http://localhost:8080/subscribe

curl -XPOST http://localhost:8080/notify-all