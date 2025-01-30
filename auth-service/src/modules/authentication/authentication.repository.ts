import {
  PasswordReset,
  Prisma,
  PrismaClient,
  Session,
  User,
} from '@prisma/client';

const prisma = new PrismaClient();

export class AuthenticationRepository {
  // User

  async createUser(data: Prisma.UserCreateInput): Promise<User> {
    try {
      return prisma.user.create({ data });
    } catch (error) {
      throw new Error(error);
    }
  }

  async getUsers(): Promise<User[]> {
    try {
      return prisma.user.findMany();
    } catch (error) {
      throw new Error(error);
    }
  }

  async getUserByEmail(email: string): Promise<User | null> {
    try {
      return prisma.user.findUnique({ where: { email } });
    } catch (error) {
      throw new Error(error);
    }
  }

  async getUserById(id: number): Promise<User | null> {
    try {
      return prisma.user.findUnique({ where: { id } });
    } catch (error) {
      throw new Error(error);
    }
  }

  async updateUser(
    id: number,
    data: Prisma.UserUpdateInput,
  ): Promise<User | null> {
    return prisma.user.update({ where: { id }, data });
  }

  async deleteUser(id: number): Promise<User | null> {
    return prisma.user.delete({ where: { id } });
  }

  // Sessions

  async createSession(data: Prisma.SessionCreateInput): Promise<Session> {
    return prisma.session.create({ data });
  }

  async getSessionByUserId(userId: number): Promise<Session> {
    return prisma.session.findFirst({ where: { userId } });
  }

  async invalidateSession(refreshToken: string): Promise<void> {
    // Refresh the session to define "expiresAt" as Date.now()
    console.log(refreshToken)
    await prisma.session.update({
      where: { refreshToken },
      data: { expiresAt: new Date() },
    });
  }

  async updateSession(
    id: number,
    data: Prisma.SessionUpdateInput,
  ): Promise<Session> {
    return prisma.session.update({ where: { id }, data });
  }

  // Token of Password Reset

  async createPasswordReset(
    data: Prisma.PasswordResetCreateInput,
  ): Promise<PasswordReset> {
    return prisma.passwordReset.create({ data });
  }

  async getPasswordResetByToken(token: string): Promise<PasswordReset | null> {
    return prisma.passwordReset.findUnique({ where: { token } });
  }

  async deletePasswordReset(id: number): Promise<void> {
    await prisma.passwordReset.delete({ where: { id } });
  }
}
