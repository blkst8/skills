import { Controller, Get } from '@nestjs/common';
import type { DashboardSummary } from '@dashboard/shared';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  /** Response is typed with the shared contract — the web app imports the same type. */
  @Get('/summary')
  getSummary(): Promise<DashboardSummary> {
    return this.appService.getSummary();
  }
}
