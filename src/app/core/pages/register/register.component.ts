import { Component, OnInit } from '@angular/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { FormControl, FormGroup, NonNullableFormBuilder, Validators, ValidatorFn, AbstractControl } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { AuthenticationService } from '../../services/authentication.service';
import { BlobService } from '../../services/blob.service';

import { CommonModule } from '@angular/common';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzMessageModule, NzMessageService } from 'ng-zorro-antd/message';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { RouterModule } from '@angular/router';
import { IRegisterUser } from '../../../utils/interfaces/user.interface';

@Component({
  selector: 'app-register',
  standalone: true,
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.less'],
  imports: [
    CommonModule,
    NzFormModule,
    NzInputModule,
    NzMessageModule,
    NzButtonModule,
    NzSpinModule,
    ReactiveFormsModule,
    TranslateModule,
    RouterModule
  ]
})
export class RegisterComponent implements OnInit {
  [key: string]: any;
  public nameErrorTip: string = '';
  public namePlaceholder: string = '';
  public emailErrorTip: string = '';
  public emailPlaceholder: string = '';
  public passwordErrorTip: string = '';
  public passwordPlaceholder: string = '';
  public confirmPasswordErrorTip: string = '';
  public confirmPasswordPlaceholder: string = '';
  public safeLogoImageUrl: SafeUrl = '';
  public isLoading: boolean = false;

  public registerForm: FormGroup<{
    name: FormControl<string>;
    email: FormControl<string>;
    password: FormControl<string>;
    confirmPassword: FormControl<string>;
  }> = this.fb.group({
    name: ['', [Validators.required]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(1)]],
    confirmPassword: ['', [Validators.required, this.matchPasswordValidator()]],
  });

  private translations = {
    'REGISTER.NAME': 'namePlaceholder',
    'REGISTER.EMAIL': 'emailPlaceholder',
    'REGISTER.PASSWORD': 'passwordPlaceholder',
    'REGISTER.CONFIRM_PASSWORD': 'confirmPasswordPlaceholder',
    'REGISTER.NAME_REQUIRED': 'nameErrorTip',
    'REGISTER.EMAIL_REQUIRED': 'emailErrorTip',
    'REGISTER.PASSWORD_REQUIRED': 'passwordErrorTip',
    'REGISTER.CONFIRM_PASSWORD_REQUIRED': 'confirmPasswordErrorTip',
    'REGISTER.PASSWORD_MISMATCH': 'confirmPasswordErrorTip'
  };
  private logoUrl: string = '../../../../assets/images/logo-black.png';

  constructor(
    private readonly fb: NonNullableFormBuilder,
    private readonly translateService: TranslateService,
    private readonly blobService: BlobService,
    private readonly sanitizer: DomSanitizer,
    private readonly authService: AuthenticationService,
    private readonly message: NzMessageService
  ) { }

  ngOnInit() {
    this.blobService.getBase64FromUrl(this.logoUrl).subscribe((base64String: string) => {
      this.safeLogoImageUrl = this.sanitizer.bypassSecurityTrustUrl('data:image/png;base64,' + base64String);
    });

    Object.entries(this.translations).forEach(([key, value]) => {
      this.translateService.get(key).subscribe((res: string) => {
        this[value] = res;
      });
    });
  }

  private matchPasswordValidator(): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } | null => {
      if (!this.registerForm) {
        return null;
      }
      const password = this.registerForm.get('password')?.value;
      const confirmPassword = control.value;
      return password === confirmPassword ? null : { 'passwordMismatch': true };
    };
  }

  submitRegisterForm(): void {
    if (this.registerForm.valid) {
      delete this.registerForm.value.confirmPassword;
      this.isLoading = true;

      this.authService.register(this.registerForm.value as IRegisterUser).subscribe((response: any) => {}).add(() => {
        this.isLoading = false
        this.createMessage('success', 'REGISTER.SUCCESS');
      });
    } else {
      Object.values(this.registerForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  createMessage(type: string, message: string): void {
    this.message.create(type, this.translateService.instant(message));
  }
}