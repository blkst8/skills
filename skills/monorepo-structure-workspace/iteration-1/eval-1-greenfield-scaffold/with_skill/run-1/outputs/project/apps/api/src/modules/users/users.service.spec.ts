import { UsersService } from "./users.service";

describe("UsersService", () => {
  it("creates a user with a generated id and the default role", () => {
    const service = new UsersService();

    const user = service.create({ email: "ada@example.com", name: "Ada" });

    expect(user.id).toBeDefined();
    expect(user.email).toBe("ada@example.com");
    expect(user.role).toBe("member");
    expect(typeof user.createdAt).toBe("string");
  });

  it("respects an explicit role", () => {
    const service = new UsersService();

    const user = service.create({ email: "grace@example.com", name: "Grace", role: "admin" });

    expect(user.role).toBe("admin");
  });

  it("lists the users that were created", () => {
    const service = new UsersService();
    service.create({ email: "ada@example.com", name: "Ada" });

    expect(service.list()).toHaveLength(1);
  });
});
