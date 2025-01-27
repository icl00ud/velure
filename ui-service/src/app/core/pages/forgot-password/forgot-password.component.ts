import { Component } from '@angular/core';
import { FormControl, FormGroup, NonNullableFormBuilder, Validators } from '@angular/forms';
import { AuthenticationService } from '../../services/authentication.service';

import { ReactiveFormsModule } from '@angular/forms';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { CommonModule } from '@angular/common';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { BlobService } from '../../services/blob.service';

@Component({
  selector: 'app-forgot-password',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzButtonModule,
    TranslateModule,
    NzFormModule,
    NzInputModule,
    NzSpinModule,
  ],
  templateUrl: './forgot-password.component.html',
  styleUrls: ['./forgot-password.component.less']
})
export class ForgotPasswordComponent {
  public emailPlaceholder: string = '';
  public emailErrorTip: string = '';
  public safeLogoImageUrl: SafeUrl = '';
  public isLoading: boolean = false;
  private logoUrl: string = '../../../../assets/images/logo-black.png';

  public forgotPasswordForm: FormGroup<{
    email: FormControl<string>;
  }>;

  constructor(
    private fb: NonNullableFormBuilder,
    private translateService: TranslateService,
    private blobService: BlobService,
    private sanitizer: DomSanitizer,
    private readonly authService: AuthenticationService,
  ) {
    this.forgotPasswordForm = this.fb.group({
      email: ['', [Validators.required]],
    });
  }

  ngOnInit() {
    this.translateService.get('FORGOT_PASSWORD.EMAIL_PLACEHOLDER').subscribe((res: string) => {
      this.emailPlaceholder = res;
    });

    this.translateService.get('FORGOT_PASSWORD.EMAIL_REQUIRED').subscribe((res: string) => {
      this.emailErrorTip = res;
    });

    this.blobService.getBase64FromUrl(this.logoUrl).subscribe((base64String: string) => this.safeLogoImageUrl = this.sanitizer.bypassSecurityTrustUrl('data:image/png;base64,' + base64String));
  }

  submitForm(): void {
    if (this.forgotPasswordForm.valid) {
      this.isLoading = true;

      //const email: string = this.forgotPasswordForm.value.email;
      
      // Lógica para solicitar a recuperação de senha
      // this.authService.forgotPassword(email).subscribe(() => {
      //   // Lógica para lidar com o sucesso da solicitação de recuperação de senha
      //   console.log('Solicitação de recuperação de senha enviada com sucesso');
      // }).add(() => {
      //   this.isLoading = false;
      // });
    } else {
      Object.values(this.forgotPasswordForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }
}
