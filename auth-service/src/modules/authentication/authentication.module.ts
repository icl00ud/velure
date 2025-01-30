import { Module } from '@nestjs/common';
import { AuthenticationService } from './authentication.service';
import { AuthenticationController } from './authentication.controller';
import { AuthenticationRepository } from './authentication.repository';
import { JwtModule } from '@nestjs/jwt';

@Module({
  imports: [JwtModule],
  controllers: [AuthenticationController],
  providers: [AuthenticationService, AuthenticationRepository],
})
export class AuthenticationModule { }
