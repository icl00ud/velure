import { Component, OnInit } from '@angular/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { FormControl, FormGroup, NonNullableFormBuilder, Validators } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { AuthenticationService } from '../../services/authentication.service';

import { CommonModule } from '@angular/common';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { Router, RouterModule } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';

import { BlobService } from '../../services/blob.service';
import { ILoginResponse, ILoginUser } from '../../../utils/interfaces/user.interface';
import { finalize } from 'rxjs/operators';

@Component({
  selector: 'app-login',
  standalone: true,
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.less'],
  imports: [
    CommonModule,
    NzSpinModule,
    NzFormModule,
    NzInputModule,
    NzButtonModule,
    NzAlertModule,
    ReactiveFormsModule,
    RouterModule,
    TranslateModule,
  ]
})
export class LoginComponent implements OnInit {
  [key: string]: any;
  public passwordErrorTip: string = '';
  public passwordPlaceholder: string = '';
  public userPlaceholder: string = '';
  public userErrorTip: string = '';
  public safeLogoImageUrl: SafeUrl = '';
  public isLoading: boolean = false;

  public validateForm: FormGroup<{
    email: FormControl<string>;
    password: FormControl<string>;
  }>;

  public errorMessage: string | null = null;

  private translations = {
    'LOGIN.USER_EMAIL': 'userPlaceholder',
    'LOGIN.PASSWORD': 'passwordPlaceholder',
    'LOGIN.USER_EMAIL_REQUIRED': 'userErrorTip',
    'LOGIN.PASSWORD_REQUIRED': 'passwordErrorTip'
  };
  private logoUrl: string = '../../../../assets/images/logo-black.png';

  constructor(
    private readonly fb: NonNullableFormBuilder,
    private readonly translateService: TranslateService,
    private readonly blobService: BlobService,
    private readonly sanitizer: DomSanitizer,
    private readonly authService: AuthenticationService,
    private readonly router: Router,
  ) {
    this.validateForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required]],
    });
  }

  ngOnInit(): void {
    this.blobService.getBase64FromUrl(this.logoUrl).subscribe((base64String: string) => {
      this.safeLogoImageUrl = this.sanitizer.bypassSecurityTrustUrl('data:image/png;base64,' + base64String);
    });

    Object.entries(this.translations).forEach(([key, value]) => {
      this.translateService.get(key).subscribe((res: string) => {
        this[value] = res;
      });
    });
  }

  submitForm(): void {
    if (this.validateForm.valid) {
      this.isLoading = true;
      this.errorMessage = null;
      const loginUserData = this.validateForm.value as ILoginUser;

      this.authService.login(loginUserData)
        .pipe(
          finalize(() => {
            this.isLoading = false;
          })
        )
        .subscribe(
          (response: ILoginResponse) => {
            localStorage.setItem('token', JSON.stringify(response));
            this.router.navigate(['/home']);
          },
          (error) => {
            if (error.status === 400 && error.error && error.error.message) {
              this.translateService.get('LOGIN.INVALID_CREDENTIALS').subscribe((res: string) => {
                this.errorMessage = res || 'Credenciais inválidas. Por favor, tente novamente.';
              });
            } else {
              this.translateService.get('LOGIN.UNEXPECTED_ERROR').subscribe((res: string) => {
                this.errorMessage = res || 'Ocorreu um erro inesperado. Por favor, tente novamente mais tarde.';
              });
            }
          }
        );
    } else {
      this.errorMessage = null;

      Object.values(this.validateForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }
}
