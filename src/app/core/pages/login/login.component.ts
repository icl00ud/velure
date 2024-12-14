import { Component } from '@angular/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { FormControl, FormGroup, NonNullableFormBuilder, Validators } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { AuthenticationService } from '../../services/authentication.service';

import { CommonModule } from '@angular/common';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { Router, RouterModule } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';

import { BlobService } from '../../services/blob.service';
import { ILoginResponse, ILoginUser } from '../../../utils/interfaces/user.interface';
import { ConfigService } from '../../config/config.service';

@Component({
  selector: 'app-login',
  standalone: true,
  templateUrl: './login.component.html',
  styleUrl: './login.component.less',
  imports: [
    CommonModule,
    NzSpinModule,
    NzFormModule,
    NzInputModule,
    NzButtonModule,
    ReactiveFormsModule,
    RouterModule,
    TranslateModule,
  ]
})
export class LoginComponent {
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
  }> = this.fb.group({
    email: ['', [Validators.required]],
    password: ['', [Validators.required]],
  });

  private translations = {
    'LOGIN.USER_EMAIL': 'userPlaceholder',
    'LOGIN.PASSWORD': 'passwordPlaceholder',
    'LOGIN.USER_EMAIL_REQUIRED': 'userErrorTip',
    'LOGIN.PASSWORD_REQUIRED': 'passwordErrorTip'
  };
  private logoUrl: string = '../../../../assets/images/logo-black.png';
  public urls = {};

  constructor(
    private readonly fb: NonNullableFormBuilder,
    private readonly translateService: TranslateService,
    private readonly blobService: BlobService,
    private readonly sanitizer: DomSanitizer,
    private readonly authService: AuthenticationService,
    private readonly router: Router,
    private readonly config: ConfigService
  ) { }

  ngOnInit() {
    this.blobService.getBase64FromUrl(this.logoUrl).subscribe((base64String: string) => this.safeLogoImageUrl = this.sanitizer.bypassSecurityTrustUrl('data:image/png;base64,' + base64String));
    this.urls = this.config.getProductServiceUrl();

    Object.entries(this.translations).forEach(([key, value]) => {
      this.translateService.get(key).subscribe((res: string) => {
        this[value] = res;
      });
    });
  }

  submitForm(): void {
    if (this.validateForm.valid) {
      this.isLoading = true;
      const loginUserData = this.validateForm.value as ILoginUser;

      this.authService.login(loginUserData).subscribe((response: ILoginResponse) => {
        localStorage.setItem('token', JSON.stringify(response));
      }).add(() => {
        this.isLoading = false;
        this.router.navigate(['/home']);
      });
    } else {
      Object.values(this.validateForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }
}
