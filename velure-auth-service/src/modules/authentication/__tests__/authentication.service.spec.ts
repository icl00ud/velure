import { Test, TestingModule } from '@nestjs/testing';
import { AuthenticationService } from '../authentication.service';
import { AuthenticationRepository } from '../authentication.repository';
import { JwtService } from '@nestjs/jwt';
import { BadRequestException } from '@nestjs/common';
import * as bcrypt from 'bcrypt';
import { CreateAuthenticationDto } from '../dto/create-authentication.dto';
import { ILoginResponse } from '../dto/login-response-dto';
import { Prisma, Session, User } from '@prisma/client';
import { ConfigService } from '@nestjs/config';

jest.mock('../authentication.repository');
jest.mock('@nestjs/jwt');
jest.mock('@nestjs/config');

describe('AuthenticationService', () => {
  let service: AuthenticationService;
  let repository: AuthenticationRepository;
  let jwtService: JwtService;
  let configService: ConfigService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthenticationService,
        AuthenticationRepository,
        JwtService,
        ConfigService,
      ],
    }).compile();

    service = module.get<AuthenticationService>(AuthenticationService);
    repository = module.get<AuthenticationRepository>(AuthenticationRepository);
    jwtService = module.get<JwtService>(JwtService);
    configService = module.get<ConfigService>(ConfigService);
  });

  it('should be defined', () => {
    expect(service as any).toBeDefined();
  });

  describe('createUser', () => {
    it('should throw an error if user already exists', async () => {
      const createUserDto = { email: 'test@example.com', password: 'password' } as CreateAuthenticationDto;
      jest.spyOn(service, 'getUserByEmail').mockResolvedValueOnce({} as User);

      await expect(service.createUser(createUserDto)).rejects.toThrowError(BadRequestException);
    });

    it('should create a new user', async () => {
      const createUserDto = { email: 'test@example.com', password: 'password' } as CreateAuthenticationDto;
      jest.spyOn(service, 'getUserByEmail').mockResolvedValueOnce(null);
      jest.spyOn(bcrypt, 'hash').mockResolvedValueOnce('hashedPassword');
      jest.spyOn(repository, 'createUser').mockResolvedValueOnce({} as User);

      const result = await service.createUser(createUserDto);

      expect(repository.createUser).toHaveBeenCalled();
      expect(result).toBeDefined();
    });
  });

  describe('getUsers', () => {
    it('should return all users without passwords', async () => {
      const users = [{ id: 1, email: 'test@example.com', password: 'password' } as User];
      jest.spyOn(repository, 'getUsers').mockResolvedValueOnce(users);

      const result = await service.getUsers();

      expect(result[0].password).toBeUndefined();
    });
  });

  describe('getUserById', () => {
    it('should return a user without password', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password' } as User;
      jest.spyOn(repository, 'getUserById').mockResolvedValueOnce(user);

      const result = await service.getUserById(1);

      expect(result.password).toBeUndefined();
    });

    it('should return null if user not found', async () => {
      jest.spyOn(repository, 'getUserById').mockResolvedValueOnce(null);

      const result = await service.getUserById(1);

      expect(result).toBeNull();
    });
  });

  describe('getUserByEmail', () => {
    it('should return a user by email', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password' } as User;
      jest.spyOn(repository, 'getUserByEmail').mockResolvedValueOnce(user);

      const result = await service.getUserByEmail('test@example.com');

      expect(result).toEqual(user);
    });

    it('should return null if user not found', async () => {
      jest.spyOn(repository, 'getUserByEmail').mockResolvedValueOnce(null);

      const result = await service.getUserByEmail('test@example.com');

      expect(result).toBeNull();
    });
  });

  describe('login', () => {
    it('should throw an error if user not found', async () => {
      jest.spyOn(service, 'getUserByEmail').mockResolvedValueOnce(null);

      await expect(service.login('test@example.com', 'password')).rejects.toThrowError('Invalid credentials');
    });

    it('should throw an error if password does not match', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'hashedPassword' } as User;
      jest.spyOn(service, 'getUserByEmail').mockResolvedValueOnce(user);
      jest.spyOn(bcrypt, 'compare').mockResolvedValueOnce(false);

      await expect(service.login('test@example.com', 'password')).rejects.toThrowError('Invalid credentials');
    });

    it('should return access and refresh tokens on successful login', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'hashedPassword' } as User;
      const session = { accessToken: 'accessToken', refreshToken: 'refreshToken' } as Session;
      jest.spyOn(service, 'getUserByEmail').mockResolvedValueOnce(user);
      jest.spyOn(bcrypt, 'compare').mockResolvedValueOnce(true);
      jest.spyOn(service as any, 'updateOrCreateSession').mockResolvedValueOnce(session);

      const result = await service.login('test@example.com', 'password');

      expect(result).toEqual({ accessToken: 'accessToken', refreshToken: 'refreshToken' });
    });
  });

  describe('logout', () => {
    it('should invalidate the session', async () => {
      jest.spyOn(repository, 'invalidateSession').mockResolvedValueOnce();

      await service.logout('refreshToken');

      expect(repository.invalidateSession).toHaveBeenCalledWith('refreshToken');
    });
  });

  describe('validateAccessToken', () => {
    it('should return a user if token is valid', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password' } as User;
      const payload = { userId: 1 };
      jest.spyOn(jwtService, 'verify').mockReturnValue(payload);
      jest.spyOn(service, 'getUserById').mockResolvedValueOnce(user);

      const result = await service.validateAccessToken('token');

      expect(result).toEqual(user);
    });

    it('should throw an error if token is invalid', async () => {
      jest.spyOn(jwtService, 'verify').mockImplementation(() => {
        throw new Error('Invalid token');
      });

      await expect(service.validateAccessToken('token')).rejects.toThrowError('Invalid token');
    });
  });

  describe('updateOrCreateSession', () => {
    it('should update an existing session', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password' } as User;
      const session = { id: 1, userId: 1, accessToken: 'accessToken', refreshToken: 'refreshToken', expiresAt: new Date() } as Session;
      jest.spyOn(service, 'getUserById').mockResolvedValueOnce(user);
      jest.spyOn(service as any, 'generateAccessToken').mockResolvedValueOnce('newAccessToken');
      jest.spyOn(service as any, 'generateRefreshToken').mockResolvedValueOnce('newRefreshToken');
      jest.spyOn(repository, 'getSessionByUserId').mockResolvedValueOnce(session);
      jest.spyOn(repository, 'updateSession').mockResolvedValueOnce({ ...session, accessToken: 'newAccessToken', refreshToken: 'newRefreshToken' });

      const result = await (service as any).updateOrCreateSession('1');

      expect(result.accessToken).toEqual('newAccessToken');
      expect(result.refreshToken).toEqual('newRefreshToken');
    });

    it('should create a new session if none exists', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password' } as User;
      const session = { id: 1, userId: 1, accessToken: 'accessToken', refreshToken: 'refreshToken', expiresAt: new Date() } as Session;
      jest.spyOn(service, 'getUserById').mockResolvedValueOnce(user);
      jest.spyOn(service as any, 'generateAccessToken').mockResolvedValueOnce('newAccessToken');
      jest.spyOn(service as any, 'generateRefreshToken').mockResolvedValueOnce('newRefreshToken');
      jest.spyOn(repository, 'getSessionByUserId').mockResolvedValueOnce(null);
      jest.spyOn(repository, 'createSession').mockResolvedValueOnce(session);

      const result = await (service as any).updateOrCreateSession('1');

      expect(result).toEqual(session);
    });
  });

  describe('getSessionsByUserId', () => {
    it('should return sessions by user id', async () => {
      const session = { id: 1, userId: 1, accessToken: 'accessToken', refreshToken: 'refreshToken', expiresAt: new Date() } as Session;
      jest.spyOn(repository, 'getSessionByUserId').mockResolvedValueOnce(session);

      const result = await service.getSessionsByUserId(1);

      expect(result).toEqual(session);
    });
  });

  describe('generateAccessToken', () => {
    it('should generate an access token', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password', name: 'Test User' } as User;
      const token = 'accessToken';
      jest.spyOn(jwtService, 'sign').mockReturnValue(token);

      const result = await (service as any).generateAccessToken(user);

      expect(result).toEqual(token);
      expect(jwtService.sign).toHaveBeenCalledWith({ userId: user.id, email: user.email, role: user.name }, { secret: service['secret'] });
    });
  });

  describe('generateRefreshToken', () => {
    it('should generate a refresh token', async () => {
      const user = { id: 1, email: 'test@example.com', password: 'password', name: 'Test User' } as User;
      const token = 'refreshToken';
      jest.spyOn(jwtService, 'sign').mockReturnValue(token);
      jest.spyOn(configService, 'get').mockReturnValueOnce('refreshSecret').mockReturnValueOnce('refreshExpiresIn');

      const result = await (service as any).generateRefreshToken(user);

      expect(result).toEqual(token);
      expect(jwtService.sign).toHaveBeenCalledWith(
        { userId: user.id, email: user.email, role: user.name },
        { secret: service['secret'] + 'refreshSecret', expiresIn: 'refreshExpiresIn' }
      );
    });
  });
});