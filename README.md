# Centralized Authentication for Microservices using Traefik, Keycloak, and Golang Echo

โปรเจกต์นี้แสดงตัวอย่างการทำ Centralized Authentication สำหรับ Microservices โดยใช้ Traefik เป็น API Gateway, Keycloak สำหรับการยืนยันตัวตน, และ Golang Echo สำหรับการให้บริการในส่วนของแอปพลิเคชัน (App Service) และการยืนยันตัวตน (Auth Service). ในโปรเจกต์นี้จะมีการตั้งค่า Forward Authentication ซึ่ง Traefik จะส่งคำขอไปยัง Auth Service เพื่อทำการตรวจสอบผู้ใช้ก่อนที่จะส่งคำขอต่อไปยัง App Service.

## คุณสมบัติของโปรเจกต์

- **Traefik**: ทำหน้าที่เป็น API Gateway ที่จัดการการ Routing และ Forward Authentication
- **Keycloak**: ระบบ Identity and Access Management (IAM) สำหรับการจัดการการยืนยันตัวตนและการอนุญาต
- **Golang Echo**: เว็บเฟรมเวิร์กที่ใช้สำหรับสร้าง App Service และ Auth Service
- **Forward Authentication**: ใช้ Auth Service ในการตรวจสอบผู้ใช้และส่ง Header `X-Auth-User`, `X-Auth-User-ID`ไปที่ App Service

## โครงสร้างโปรเจกต์
```bash
.
├── README.md
├── app-service
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── auth-service
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── docker-compose.yml
├── keycloak
│   └── realm-export.json
└── traefik
    ├── dynamic_conf.yml
    └── traefik.yml
```

## วิธีrun โปรเจกต์

```bash
docker-compose up --build -d
```

## หมายเหตุ
1. ต้องไป สร้าง user ใน Keycloak หลังจากrun โปรเจกต์ เพื่อใช้งาน
2. ต้องไป เอา public key ใน Keycloak หน้า Relam Setting > Key > RSA  ให้กับ Auth Service และใส่ค่าใน publicKeyStr หลังจากrun โปรเจกต์ เพื่อใช้งาน
3. build auth service ใหม่หลังจากrun โปรเจกต์ เพื่อใช้งาน
```bash
docker-compose up --build -d auth-service
```

## คำสั่งนี้จะทำให้บริการต่อไปนี้เริ่มทำงาน:

Keycloak: ระบบจัดการผู้ใช้และการยืนยันตัวตนที่ http://keycloak.localhost:8080
Traefik: Dashboard ที่จัดการการทำงานของบริการต่างๆ ที่ http://localhost:8082
App Service: บริการแอปพลิเคชันที่เข้าถึงได้ผ่าน Traefik ที่ http://localhost/app/*
Auth Service: บริการยืนยันตัวตนที่เข้าถึงได้ผ่าน Traefik ที่ http://localhost/auth/*

## คำสั่งนี้จะทำการทำงานของระบบ
Forward Authentication: เมื่อผู้ใช้เข้าถึง http://localhost/app/* Traefik จะส่งคำขอไปยัง Auth Service ที่ http://localhost/auth/* ทำการตรวจสอบความถูกต้องของผู้ใช้. หากการตรวจสอบสำเร็จ Auth Service จะส่งค่า `X-Auth-User`,`X-Auth-User-ID`  กลับมาให้ Traefik ซึ่งจะส่งต่อให้กับ App Service
App Service: เมื่อ App Service ได้รับคำขอพร้อมกับ Header `X-Auth-User`, `X-Auth-User-ID` จะใช้ข้อมูลนี้ในการดำเนินการภายในแอปพลิเคชัน เช่น การแสดงข้อมูลส่วนบุคคลหรือการดำเนินการอื่นๆ ที่เกี่ยวข้องกับผู้ใช้
