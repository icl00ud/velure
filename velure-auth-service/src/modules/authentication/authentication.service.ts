import { BadRequestException, Injectable } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import * as bcrypt from 'bcrypt';
import { CreateAuthenticationDto } from './dto/create-authentication.dto';
import { AuthenticationRepository } from './authentication.repository';
import { Prisma, Session, User } from '@prisma/client';
import { ILoginResponse } from './dto/login-response-dto';

@Injectable()
export class AuthenticationService {
  private readonly saltRounds = 10;
  private readonly secret = this.configService.get<string>('jwt.secret');

  constructor(
    private readonly authenticationRepository: AuthenticationRepository,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
  ) {}

  async createUser(data: CreateAuthenticationDto): Promise<User> {
    const existingUser = await this.getUserByEmail(data.email);

    if (existingUser) {
      throw new BadRequestException('User already exists');
    }

    const userData: CreateAuthenticationDto = data;
    const hashedPassword = await bcrypt.hash(data.password, this.saltRounds);
    userData.password = hashedPassword;

    const newUser = await this.authenticationRepository.createUser(userData);
    return newUser;
  }

  async getUsers(): Promise<User[]> {
    const users = await this.authenticationRepository.getUsers();

    for (const user of users) {
      delete user.password;
    }

    return users;
  }

  async getUserById(id: number): Promise<User | null> {
    const user = await this.authenticationRepository.getUserById(id);

    if (user) {
      delete user.password;
    }

    return user;
  }

  async getUserByEmail(email: string): Promise<User | null> {
    return this.authenticationRepository.getUserByEmail(email);
  }

  async login(email: string, password: string): Promise<ILoginResponse> {
    const user = await this.getUserByEmail(email);

    if (!user) {
      throw new Error('Invalid credentials');
    }

    const passwordMatches = await bcrypt.compare(password, user.password);

    if (!passwordMatches) {
      throw new Error('Invalid credentials');
    }

    // Atualização da sessão ao fazer login
    const session = await this.updateOrCreateSession(user.id.toString());
    return {
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
    };
  }

  async logout(refreshToken: string): Promise<void> {
    await this.authenticationRepository.invalidateSession(refreshToken);
  }

  private async updateOrCreateSession(userId: string): Promise<Session> {
    const sessionExpiresIn: string = this.configService.get<string>('session.expiresIn');
    const user: User = await this.getUserById(parseInt(userId));
    const [accessToken, refreshToken] = await Promise.all([
      this.generateAccessToken(user),
      this.generateRefreshToken(user),
    ]);

    let session = await this.authenticationRepository.getSessionByUserId(+userId);

    if (session) {
      session.accessToken = accessToken;
      session.refreshToken = refreshToken;
      session.expiresAt = new Date(Date.now() + parseInt(sessionExpiresIn, 10));
      session = await this.authenticationRepository.updateSession(session.id, session);
    } else {
      const sessionData: Prisma.SessionCreateInput = {
        accessToken,
        refreshToken,
        expiresAt: new Date(Date.now() + parseInt(sessionExpiresIn, 10)),
        user: {
          connect: { id: parseInt(userId) },
        },
      };
      session = await this.authenticationRepository.createSession(sessionData);
    }

    return session;
  }

  async validateAccessToken(token: string): Promise<User> {
    try {
      const payload = this.jwtService.verify(token, { secret: this.secret });
      const user = await this.getUserById(payload.userId);
      if (!user) throw new Error('User not found');
      return user;
    } catch (error) {
      console.error(error);
      throw new Error('Invalid token');
    }
  }

  async getSessionsByUserId(userId: number): Promise<Session> {
    return this.authenticationRepository.getSessionByUserId(userId);
  }

  private async generateAccessToken(user: User): Promise<string> {
    const payload = { userId: user.id, email: user.email, role: user.name };
    return this.jwtService.sign(payload, { secret: this.secret });
  }

  private async generateRefreshToken(user: User): Promise<string> {
    const refreshSecret: string = this.configService.get<string>('jwt.refreshSecret');
    const refreshExpiresIn: string = this.configService.get<string>('jwt.refreshExpiresIn');

    const payload = { userId: user.id, email: user.email, role: user.name };
    const secret = this.secret + refreshSecret;
    return this.jwtService.sign(payload, {secret, expiresIn: refreshExpiresIn });
  }
}
