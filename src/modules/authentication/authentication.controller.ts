import { Controller, Get, Post, Body, Param, Delete } from '@nestjs/common';

import { AuthenticationService } from './authentication.service';

import { CreateAuthenticationDto } from './dto/create-authentication.dto';

import { User } from '@prisma/client';
import { ILoginResponse } from './dto/login-response-dto';

@Controller('authentication')
export class AuthenticationController {
  constructor(private readonly authService: AuthenticationService) { }

  @Post('register')
  async register(@Body() createAuthDto: CreateAuthenticationDto,): Promise<CreateAuthenticationDto> {
    try {
      return await this.authService.createUser(createAuthDto);
    } catch (error) {
      throw new Error(error);
    }
  }

  @Post('login')
  async login(@Body('email') email: string, @Body('password') password: string): Promise<ILoginResponse> {
    return await this.authService.login(email, password);
  }

  @Post('validateToken')
  async validateToken(@Body('token') token: string,): Promise<{ isValid: boolean }> {
    try {
      const user = await this.authService.validateAccessToken(token);
      return { isValid: !!user };
    } catch (error) {
      return { isValid: false };
    }
  }

  @Get('users')
  async getUsers(): Promise<User[]> {
    return await this.authService.getUsers();
  }

  @Get('user/id/:id')
  async getUserById(@Param('id') id: string): Promise<User> {
    return await this.authService.getUserById(+id);
  }

  @Get('user/email/:email')
  async getUserByEmail(@Param('email') email: string): Promise<User> {
    return await this.authService.getUserByEmail(email);
  }

  @Post('logout')
  async logout(@Body() refreshToken: string): Promise<void> {
    await this.authService.logout(refreshToken);
  }
}
