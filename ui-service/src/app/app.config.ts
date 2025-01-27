import { ApplicationConfig, importProvidersFrom } from '@angular/core';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideRouter } from '@angular/router';

import { TranslateLoader, TranslateModule } from '@ngx-translate/core';
import { TranslateService } from './core/services/translate.service';

import { routes } from './app.routes';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { DashboardOutline, MenuUnfoldOutline, FormOutline, MenuFoldOutline, ShoppingCartOutline, UserOutline, GithubFill, GithubOutline, LockOutline, MailOutline, DeleteOutline, DeleteFill, DeleteTwoTone } from '@ant-design/icons-angular/icons';

const icons = [
  MenuFoldOutline,
  MenuUnfoldOutline,
  DashboardOutline,
  FormOutline,
  ShoppingCartOutline,
  UserOutline,
  GithubOutline,
  LockOutline,
  MailOutline,
  DeleteOutline,
  DeleteFill,
  DeleteTwoTone
];

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideAnimations(),
    importProvidersFrom(
      HttpClientModule,
      NzIconModule.forRoot(icons),
      TranslateModule.forRoot({
        defaultLanguage: 'pt',
        useDefaultLang: true,
        loader: {
          provide: TranslateLoader,
          useClass: TranslateService,
          deps: [HttpClient]
        }
      }),
    )
  ]
};