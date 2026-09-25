import { Controller, Get } from "@nestjs/common";
import { API_ROUTES } from "@repo/shared";
import type { HealthCheckDto } from "@repo/types";

@Controller(API_ROUTES.health)
export class HealthController {
  @Get()
  check(): HealthCheckDto {
    return {
      status: "ok",
      uptime: process.uptime(),
      timestamp: new Date().toISOString(),
    };
  }
}
