package com.banka1.banking.aspect;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Authorization {
    boolean customerOnlyOperation() default false;
    boolean employeeOnlyOperation() default false;
    boolean disallowAdminFallback() default false;
}
