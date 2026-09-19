import { Injectable } from '@nestjs/common';
import { PrismaClient } from '@dashboard/database';
import type { DashboardSummary } from '@dashboard/shared';

@Injectable()
export class AppService {
  constructor(private readonly prisma: PrismaClient) {}

  async getSummary(): Promise<DashboardSummary> {
    const totalWidgets = await this.prisma.widget.count();
    const latest = await this.prisma.widget.findFirst({
      orderBy: { updatedAt: 'desc' },
      select: { updatedAt: true },
    });

    return {
      totalWidgets,
      lastUpdated: latest?.updatedAt.toISOString() ?? new Date(0).toISOString(),
    };
  }
}
