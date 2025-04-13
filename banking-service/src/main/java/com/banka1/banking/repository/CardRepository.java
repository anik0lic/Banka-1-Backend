package com.banka1.banking.repository;

import com.banka1.banking.models.Card;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface CardRepository extends JpaRepository<Card, Long> {

    Optional<List<Card>> findByAccountId(Long accountID);

    Optional<List<Card>> findByAccountIdAndActive(Long account_id, Boolean active);
}
