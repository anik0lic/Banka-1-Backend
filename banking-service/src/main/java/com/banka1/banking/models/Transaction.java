package com.banka1.banking.models;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

@Entity
@Getter
@Setter
public class Transaction {
    @Id
    private Long id;

    @ManyToOne
    @JoinColumn(name = "from_account_id", nullable = false)
    private Account fromAccountId;

    @ManyToOne
    @JoinColumn(name = "to_account_id", nullable = false)
    private Account toAccountId;

    @Column(nullable = false)
    private Double amount;

    @ManyToOne
    @JoinColumn(name = "currency_id", nullable = false)
    private Currency currency;

    @Column(nullable = false)
    private Long timestamp;

    @Column(nullable = true)
    private String description;

    @ManyToOne
    @JoinColumn(name = "transfer_id", nullable = false)
    private Transfer transfer;

    private Long loanId; // ukoliko je uplata za kredit, mora biti postavljen id kredita
}
