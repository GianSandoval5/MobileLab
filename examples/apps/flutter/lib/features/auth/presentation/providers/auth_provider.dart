import 'package:app_prueba_mobilelab/core/providers/repository_providers.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/repositories/auth_repository.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/usecases/login.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/usecases/register.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class AuthState {
  const AuthState({this.user, this.isLoading = false, this.error});

  final AppUser? user;
  final bool isLoading;
  final String? error;
}

class AuthController extends StateNotifier<AuthState> {
  AuthController(this._login, this._register, this._repository)
    : super(const AuthState());

  final Login _login;
  final Register _register;
  final AuthRepository _repository;

  Future<bool> login({required String email, required String password}) async {
    state = const AuthState(isLoading: true);
    try {
      final session = await _login(email: email, password: password);
      state = AuthState(user: session.user);
      return true;
    } catch (error) {
      state = AuthState(error: error.toString());
      return false;
    }
  }

  Future<bool> register({
    required String name,
    required String email,
    required String password,
  }) async {
    state = const AuthState(isLoading: true);
    try {
      final session = await _register(
        name: name,
        email: email,
        password: password,
      );
      state = AuthState(user: session.user);
      return true;
    } catch (error) {
      state = AuthState(error: error.toString());
      return false;
    }
  }

  Future<bool> updateProfile({
    required String name,
    required String email,
    required String phone,
  }) async {
    final currentUser = state.user;
    state = AuthState(user: currentUser, isLoading: true);
    try {
      final updated = await _repository.updateProfile(
        name: name,
        email: email,
        phone: phone,
      );
      // MobileLab usa una fixture fija; conservamos los valores editados en UI.
      state = AuthState(
        user: updated.copyWith(name: name, email: email, phone: phone),
      );
      return true;
    } catch (error) {
      state = AuthState(user: currentUser, error: error.toString());
      return false;
    }
  }

  void logout() => state = const AuthState();

  void clearError() {
    state = AuthState(user: state.user, isLoading: state.isLoading);
  }
}

final authControllerProvider = StateNotifierProvider<AuthController, AuthState>(
  (ref) {
    return AuthController(
      ref.watch(loginUseCaseProvider),
      ref.watch(registerUseCaseProvider),
      ref.watch(authRepositoryProvider),
    );
  },
);
